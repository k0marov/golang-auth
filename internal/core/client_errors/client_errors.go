package client_errors

type ClientError struct {
	ReadableDetail string
	DetailCode     string
}

func (ce ClientError) Error() string {
	return "An error which will be displayed to the client: " + ce.ReadableDetail
}

var InvalidJsonError = ClientError{
	DetailCode:     "invalid-json",
	ReadableDetail: "The provided request body is not valid JSON.",
}

var UnhashedPasswordError = ClientError{
	DetailCode:     "unhashed-password",
	ReadableDetail: "The provided password is not client-side hashed. Please hash it and try again (better with a new password).",
}

var UsernameAlreadyTakenError = ClientError{
	DetailCode:     "username-taken",
	ReadableDetail: "A user with that username already exists.",
}

var UsernameInvalidError = ClientError{
	DetailCode:     "username-invalid",
	ReadableDetail: "Username is invalid. Usernames can only contain latin characters, digits and underscores, and cannot start with underscore.",
}

var InvalidCredentialsError = ClientError{
	DetailCode:     "invalid-credentials",
	ReadableDetail: "Login failed: username and password don't match.",
}

var AuthTokenRequiredError = ClientError{
	DetailCode:     "token-required",
	ReadableDetail: "For accesing this resource you need to provide authentication credentials via Authorization: Token {TOKEN} header.",
}

var AuthTokenInvalidError = ClientError{
	DetailCode:     "token-invalid",
	ReadableDetail: "The Auth token you provided is invalid (maybe it has expired).",
}
