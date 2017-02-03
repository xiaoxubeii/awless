package aws

import (
	"errors"
	"sync"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/wallix/awless/cloud"
)

var (
	AccessService, InfraService cloud.Service

	SecuAPI Security
)

func InitSession(region string) (*session.Session, error) {
	session, err := session.NewSession(
		&awssdk.Config{
			Region: awssdk.String(region),
			Credentials: credentials.NewChainCredentials([]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
			})})
	if err != nil {
		return nil, err
	}
	if _, err = session.Config.Credentials.Get(); err != nil {
		return nil, errors.New(`Your AWS credentials seem undefined!
AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY need to be exported in your CLI environment

Installation documentation is at https://github.com/wallix/awless/wiki/Installation`)
	}

	return session, nil
}

func InitServices(region string) error {
	sess, err := InitSession(region)
	if err != nil {
		return err
	}
	AccessService = NewAccess(sess)
	InfraService = NewInfra(sess)
	SecuAPI = NewSecu(sess)

	return nil
}

func multiFetch(fns ...func() (interface{}, error)) (<-chan interface{}, <-chan error) {
	resultc := make(chan interface{})
	errc := make(chan error, 1)

	var wg sync.WaitGroup

	for _, fn := range fns {
		wg.Add(1)
		go func(fetchFn func() (interface{}, error)) {
			defer wg.Done()
			r, err := fetchFn()
			if err != nil {
				errc <- err
				return
			}
			resultc <- r
		}(fn)
	}

	go func() {
		wg.Wait()
		close(resultc)
		close(errc)
	}()

	return resultc, errc
}