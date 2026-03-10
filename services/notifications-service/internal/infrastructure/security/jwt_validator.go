package security

import pkgjwt "aurora/pkg/jwt"

type JWTValidator = pkgjwt.JWTValidator
type TokenClaims = pkgjwt.TokenClaims

var ErrTokenValidation = pkgjwt.ErrTokenValidation
var NewJWTValidator = pkgjwt.NewJWTValidator
