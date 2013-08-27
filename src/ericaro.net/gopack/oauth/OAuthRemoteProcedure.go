package oauth

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"ericaro.net/gopack"
	"ericaro.net/gopack/protocol"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const gpkOauthConsumerKey string = "gpkOauthConsumerKey"

const introductionTitle string = "\n----------- OAuth 1.0 Token Procedure ----------\n"
const introductionText string = "This wizzard will help you to configure OAuth \nauthentification for the new repository."

const promptPemFile string = "Where is the PEM file (containing your private RSA key)"
const promptRequestTokenUrl string = "What is the request token URL"
const promptAuthorizeTokenUrl string = "What is the authorize token URL"
const promptAccessTokenUrl string = "What is the access token URL"

const authorizationText string = "\nPlease browse the URL below, allow the connection\n and paste the verification code."
const verificationCodeText string = "Enter the verification code"

func RequestOAuthToken(name, remote string) (t *protocol.Token, err error) {

	// Print introduction text:
	gopack.TitleStyle.Printf(introductionTitle)
	fmt.Println(introductionText)
	waitForEnter()

	// Ask for RSA pem file:
	pemFile := promptUser(promptPemFile, "/home/dominique/.ssh/id_rsa")
	if !gopack.FileExists(pemFile) {
		err = fmt.Errorf("The file '%s' doesn't exist.", pemFile)
		return nil, err
	}

	// Load private key file:
	content, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	// Decode PEM bytes:
	block, _ := pem.Decode(content)
	if block == nil {
		err = fmt.Errorf("Wrong private key format")
		return nil, err
	}

	// Instantiate the private key with decoded bytes:
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	signer := RSA_SHA1_Signer{Private: privateKey}

	// Ask for OAuth request endpoint URL:
	// requestTokenUrl := promptUser(promptRequestTokenUrl, remote)
	requestTokenUrl := promptUser(promptRequestTokenUrl, "http://192.168.111.136:8080/oauth/TemporaryCredentialRequest")

	// Ask for OAuth authorize endpoint URL:
	// authorizeTokenUrl := promptUser(promptAuthorizeTokenUrl, requestTokenUrl)
	authorizeTokenUrl := promptUser(promptAuthorizeTokenUrl, "http://192.168.111.136:8080/oauth/Authorize")

	// Ask for OAuth request endpoint URL:
	// accessTokenUrl := promptUser(promptAccessTokenUrl, authorizeTokenUrl)
	accessTokenUrl := promptUser(promptAccessTokenUrl, "http://192.168.111.136:8080/oauth/TokenCredentials")
	serviceProvider := ServiceProvider{
		RequestTokenUrl:   requestTokenUrl,
		AuthorizeTokenUrl: authorizeTokenUrl,
		AccessTokenUrl:    accessTokenUrl,
	}

	// Initialize OAuth consumer:
	consumer := NewConsumer(gpkOauthConsumerKey, "don't care", serviceProvider, &signer)

	// Request a request token:
	rtoken, loginUrl, err := consumer.GetRequestTokenAndUrl("oob")
	if err != nil {
		return nil, err
	}

	// Ask for verification code:
	fmt.Println(authorizationText)
	fmt.Printf("\nURL: %s\n", loginUrl)
	verificationCode := promptUser(verificationCodeText, "")

	// Ask server for an access token:
	atoken, err := consumer.AuthorizeToken(rtoken, verificationCode)
	if err != nil {
		return nil, err
	}

	fmt.Println("AToken:", atoken)
	os.Exit(0) //TODO remove me

	return
}

func waitForEnter() {
	fmt.Print("\n[Press ENTER]")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func promptUser(text, def string) (ans string) {
	ans = def

	// Display prompt text:
	gopack.TitleStyle.Printf("\n%s ?", text)

	// Display default value: (if any)
	if len(def) > 0 {
		fmt.Printf("\nLeave empty to use default value [%s]", def)
	}

	fmt.Print("\n> ")

	// Read input:
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimRight(input, "\n")
	if len(input) > 0 {
		ans = input
	}

	return
}
