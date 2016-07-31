package iris_jose

import (
	"gopkg.in/square/go-jose.v1"
)

const (
	//Default context key
	DefaultContextKey = "jose"
)

//Config is a struct for specifying configurations options for the iris_jose middleware
type Config struct {
	// The function that will return the private ans public key to decrypt and validate the JOSE
	// Default : nil
	KeysGetter KeyFunc

	// String representing the key management algorithm
	//default value: RSA_OAEP_256
	KeyAlgorithm jose.KeyAlgorithm

	// the name of the property in the request where the user information
	// from the JOSE will be stored
	//default value: jose
	ContextKey string

	//The function that will be called when there's an error validation the token
	ErrorHandler errorHandler

	//A boolean indicating if the credentials are required or not
	//default value: false
	CredentialsOptional bool

	//A function that extract the jose from request
	// Default: FromAuthHeader ( i.e from Authorization header as bearer token )
	Extractor TokenExtractor

	// Debug flag turns on debugging output
	// Default : false
	Debug bool

	//When set, all requests with the OPTIONS method will use authentication
	// if you enable this options, you should register youre route with iris.Options(...) also
	// Default : false
	EnableAuthOnOptions bool

	// Define the signature algorithm.
	//default : RS512
	SignatureAlgorithm jose.SignatureAlgorithm

	// specify the encryption algorythm
	// Default : A256CBC_HS512
	EncryptionAlgorithm jose.ContentEncryption
}
