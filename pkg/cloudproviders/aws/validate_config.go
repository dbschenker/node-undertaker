package aws

import (
	"fmt"
	"gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker/flags"
	"net/url"
)

func (t AwsCloudProvider) ValidateConfig() error {
	if t.SqsUrl == "" {
		return fmt.Errorf("%s can't be empty", flags.AwsSqsUrlFlag)
	}

	_, err := url.ParseRequestURI(t.SqsUrl)
	if err != nil {
		return err
	}

	return nil
}
