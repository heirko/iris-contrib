package iris_jose

import (
	"fmt"
	"log"
	"strings"

	"github.com/kataras/iris"
	"gopkg.in/square/go-jose.v1"
)

//------------------------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------------
//													GENERAL														//
//------------------------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------------
// A function called whenever a, error is encountered
type errorHandler func(*iris.Context, string)

// Token Extractor is a function that takes a context as input and returns
// either a token or an error. An error should only be returned if an attemps to specify
// a token was found, but the information was somehow incorrectly formed.
// In the case where a token is simply not present, this should not be treated as an error
type TokenExtractor func(ctx *iris.Context) (string, error)

// Callback function to supply the key for verification. Used by the Parse methodbut unverified Token. This allows one to use properties
// in the Header of the token to identify which key to use
type KeyFunc func() (interface{}, interface{}, error)

// Default error Handler
func OnError(ctx *iris.Context, err string) {
	ctx.SetStatusCode(iris.StatusUnauthorized)
	ctx.SetBodyString(err)
}

//------------------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------
// 						MIDDLEWARE 						    												  //
//------------------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------
/**
/*
 * Middleware for JOSE handling
*/
type Middleware struct {
	Config Config
}

// Middleware Constructor
func New(cfg ...Config) *Middleware {
	var c Config

	if len(cfg) == 0 {
		c = Config{}
	} else {
		c = cfg[0]
	}

	if c.KeyAlgorithm == "" {
		c.KeyAlgorithm = jose.RSA_OAEP_256
	}

	if c.ContextKey == "" {
		c.ContextKey = DefaultContextKey
	}

	if c.ErrorHandler == nil {
		c.ErrorHandler = OnError
	}

	if c.Extractor == nil {
		c.Extractor = FromAuthHeader
	}

	if c.SignatureAlgorithm == "" {
		c.SignatureAlgorithm = jose.RS512
	}

	if c.EncryptionAlgorithm == "" {
		c.EncryptionAlgorithm = jose.A256CBC_HS512
	}

	return &Middleware{Config: c}
}

func (m *Middleware) logf(format string, args ...interface{}) {
	if m.Config.Debug {
		log.Printf(format, args...)
	}
}

// Get returns the user information for this client/request
func (m *Middleware) Get(ctx *iris.Context) []byte {
	return ctx.Get(m.Config.ContextKey).([]byte)
}

// Serve the middelware's action
func (m *Middleware) Serve(ctx *iris.Context) {
	err := m.CheckToken(ctx)
	if err == nil {
		ctx.Next()
	}
}

// CheckToken main functionnality, checks for token
func (m *Middleware) CheckToken(ctx *iris.Context) error {
	if !m.Config.EnableAuthOnOptions {
		if ctx.MethodString() == iris.MethodOptions {
			return nil
		}
	}

	token, err := m.Config.Extractor(ctx)

	if err != nil {
		m.logf("Error extracting JOSE: %v", err)
		m.Config.ErrorHandler(ctx, err.Error())
		return fmt.Errorf("Error extracting token: %v", err)
	} else {
		m.logf("Token extracted: %s", token)
	}
	//if token is empty
	if token == "" {
		if m.Config.CredentialsOptional {
			m.logf(" No credentials found (CredentialOptionals=true)")
			return nil
		}
		errorMsg := "Required auhtorization token not found"
		m.Config.ErrorHandler(ctx, errorMsg)
		m.logf("Error: No credentials found (CredentialsOptional=false)")
		return fmt.Errorf(errorMsg)
	}

	//now parse the token
	claim, err := m.parse(token)

	if err != nil {
		errorMsg := fmt.Sprintf("Error: Unable to parse Token : %s", token)
		m.Config.ErrorHandler(ctx, errorMsg)
		m.logf(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	ctx.Set(m.Config.ContextKey, claim)

	return nil

}

// parse the raw token and extract data.
func (m *Middleware) parse(tokenString string) ([]byte, error) {

	privateKey, publicKey, err := m.Config.KeysGetter()
	encObject, err := jose.ParseEncrypted(tokenString)

	if err != nil {
		message := fmt.Sprintf("Error: token %s is not a valid encrypted message", tokenString)
		m.logf(message)
		return nil, fmt.Errorf(message)
	}

	decrypted, err := encObject.Decrypt(privateKey)
	if err != nil {
		message := fmt.Sprintf("Error : token %s could not be decrypted", tokenString)
		m.logf(message)
		return nil, fmt.Errorf(message)
	}

	signObject, err := jose.ParseSigned(string(decrypted))
	if err != nil {
		message := fmt.Sprintf("Error: token %s is not a valid signed message", tokenString)
		m.logf(message)
		return nil, fmt.Errorf(message)
	}

	jsonStringObjectBytes, err := signObject.Verify(publicKey)
	if err != nil {
		message := fmt.Sprintf("Error verifying token %s", tokenString)
		m.logf(message)
		return nil, fmt.Errorf(message)
	}
	m.logf("Json object verified : %s", string(jsonStringObjectBytes))

	return jsonStringObjectBytes, nil

}

//create a new token with the given claim and with given encrypt and signing algs
func (m *Middleware) NewTokenWithClaim(claim interface{}) string {

	//Todo: error handling

	privateKey, publicKey, _ := m.Config.KeysGetter()

	signer, _ := jose.NewSigner(m.Config.SignatureAlgorithm, privateKey)
	encrypter, _ := jose.NewEncrypter(m.Config.KeyAlgorithm, m.Config.EncryptionAlgorithm, publicKey)

	jsonBytes, _ := jose.MarshalJSON(claim)
	m.logf("json marshalled claim : %s", jsonBytes)

	signObject, _ := signer.Sign(jsonBytes)

	serialized := signObject.FullSerialize()
	encryptObj, _ := encrypter.Encrypt([]byte(serialized))
	serialized = encryptObj.FullSerialize()

	return serialized
}

//------------------------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------------
// 						TOKEN EXTRACTORS 						  //
//------------------------------------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------------------------------------
// FromAuthHeader is a TokenExtractor that takes a given context and extracts
// the JWS or JWE from the Authorization header
func FromAuthHeader(ctx *iris.Context) (string, error) {
	authHeader := ctx.RequestHeader("Authorization")
	if authHeader == "" {
		return "", nil
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf("Authorization header format must be Bearer {token}")
	}

	return authHeaderParts[1], nil
}

// FromParameter returns a function that extracts the token
// from the specified query string parameter
func FromParameter(param string) TokenExtractor {
	return func(ctx *iris.Context) (string, error) {
		return ctx.URLParam(param), nil
	}
}

// FromFirst returns a function that runs multiple token extractors and takes
// the first token it finds
func FromFirst(extractors ...TokenExtractor) TokenExtractor {
	return func(ctx *iris.Context) (string, error) {
		for _, ex := range extractors {
			token, err := ex(ctx)
			if err != nil {
				return "", err
			}
			if token != "" {
				return token, nil
			}

		}
		return "", nil
	}
}
