package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

var (
	cdkConfig     Config
	configsDir    = "configs"
	REMINDERS_ENV = "REMINDERS_ENV"
)

func main() {
	defer jsii.Close()
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}
	cdkConfig = *cfg
	cdkConfig.Log()

	app := awscdk.NewApp(nil)
	stack := createCDKStack(app, cdkConfig.Stack.Name)
	lf := createLambdaFunction(stack)
	createApiGateway(stack, lf)
	createDynamoDB(stack)
	app.Synth(nil)
}

func createDynamoDB(stack awscdk.Stack) {
	tp := awsdynamodb.TableProps{
		TableName: jsii.String(cdkConfig.Dynamo.Name),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String(cdkConfig.Dynamo.PartitionKey),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String(cdkConfig.Dynamo.SortKey),
			Type: awsdynamodb.AttributeType_STRING,
		},
	}

	awsdynamodb.NewTable(
		stack,
		jsii.String(cdkConfig.Dynamo.Id),
		&tp,
	)
}

func createCDKStack(scope constructs.Construct, id string) awscdk.Stack {
	// account := getAccount()
	// region := getRegion(cdkConfig.Region)

	return awscdk.NewStack(
		scope,
		&id,
		&awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(cdkConfig.Account),
				Region:  jsii.String(cdkConfig.Region),
			},
		})
}

func createLambdaFunction(stack awscdk.Stack) awslambda.Function {
	managedPolicies := []awsiam.IManagedPolicy{}
	for _, v := range cdkConfig.ManagedPolicies {
		mp := awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String(v.Name), jsii.String(v.Arn))
		managedPolicies = append(managedPolicies, mp)
	}

	roleProps := awsiam.RoleProps{
		AssumedBy:       awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		ManagedPolicies: &managedPolicies,
	}

	lambdaRole := awsiam.NewRole(
		stack,
		jsii.String(cdkConfig.Lambda.ExecutionRole),
		&roleProps,
	)

	lambdaProps := awslambda.FunctionProps{
		FunctionName: jsii.String(cdkConfig.Lambda.Name),
		Architecture: awslambda.Architecture_ARM_64(),
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Role:         lambdaRole,
		Code:         awslambda.NewAssetCode(jsii.String("RemindersFunction.zip"), nil),
		Handler:      jsii.String("bootstrap"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(10)),
	}

	return awslambda.NewFunction(
		stack,
		jsii.String(cdkConfig.Lambda.Name+"ID"),
		&lambdaProps,
	)
}

func createApiGateway(stack awscdk.Stack, lf awslambda.Function) {
	apiGatewayProps := awsapigateway.RestApiProps{
		RestApiName: jsii.String(cdkConfig.ApiGateway.Name),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String(cdkConfig.ApiGateway.Stage),
		},
	}

	mo := awsapigateway.MethodOptions{}
	ro := awsapigateway.ResourceOptions{DefaultMethodOptions: &mo}
	io := &awsapigateway.LambdaIntegrationOptions{
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
	}

	li := awsapigateway.NewLambdaIntegration(lf, io)
	apiGateway := awsapigateway.NewRestApi(
		stack,
		jsii.String(cdkConfig.ApiGateway.Id),
		&apiGatewayProps,
	)

	addRestApiResources(apiGateway, ro, li)
}

func addRestApiResources(apiGateway awsapigateway.RestApi, ro awsapigateway.ResourceOptions, li awsapigateway.LambdaIntegration) {
	// "api/v1/reminders/...."
	r := apiGateway.Root().AddResource(
		jsii.String("api"),
		&ro,
	).AddResource(
		jsii.String("v1"),
		&ro,
	).AddResource(
		jsii.String("reminders"),
		&ro,
	)

	r.AddMethod(
		jsii.String("GET"),
		li,
		nil,
	)
	r.AddMethod(
		jsii.String("POST"),
		li,
		nil,
	)

	r.AddResource(
		jsii.String("status"),
		&ro,
	).AddResource(
		jsii.String("{remindersId}"),
		&ro,
	).AddMethod(
		jsii.String("PUT"),
		li,
		nil,
	)

	r.AddResource(
		jsii.String("flag"),
		&ro,
	).AddResource(
		jsii.String("{remindersId}"),
		&ro,
	).AddMethod(
		jsii.String("PUT"),
		li,
		nil,
	)

	idR := r.AddResource(
		jsii.String("{remindersId}"),
		&ro,
	)
	idR.AddMethod(
		jsii.String("GET"),
		li,
		nil,
	)
	idR.AddMethod(
		jsii.String("DELETE"),
		li,
		nil,
	)
}

//go:embed configs/*
var configs embed.FS

func readConfig() (*Config, error) {
	var config Config

	// read env from env var
	env, err := getEnv()
	if err != nil {
		return nil, err
	}
	config.Env = env

	// now read the config file based on the env
	data, err := configs.ReadFile("configs/" + config.Env + ".json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %w", err)
	}

	// unmarshall json config into struct
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %w", err)
	}

	// read region, default is set to "us-east-1"
	// default region is taken from initial config file
	// and used if no region found in env vars
	rgn, err := getRegion(config.Region)
	if err != nil {
		return nil, err
	}
	config.Region = rgn

	// read account from env vars
	acct, err := getAccount()
	if err != nil {
		return nil, err
	}
	config.Account = acct

	return &config, nil
}

func getEnv() (string, error) {
	supportedEnv := []string{"dev", "prod"}
	env, ok := os.LookupEnv(REMINDERS_ENV)

	if !ok || len(env) == 0 {
		return "", fmt.Errorf(`environment variable REMINDERS_ENV is not set: supported env are: %s`, strings.Join(supportedEnv, ` | `))
	}

	if !slices.Contains(supportedEnv, env) {
		return "", fmt.Errorf(`environment variable REMINDERS_ENV provided is %s but only supported env are: %s`, env, strings.Join(supportedEnv, ` | `))
	}

	return env, nil
}

func getAccount() (string, error) {
	account, found := os.LookupEnv("CDK_DEPLOY_ACCOUNT")
	if !found {
		account, found = os.LookupEnv("CDK_DEFAULT_ACCOUNT")
		if !found {
			return "", fmt.Errorf("no account was found in CDK_DEPLOY_ACCOUNT and CDK_DEFAULT_ACCOUNT envs")
		}
	}
	return account, nil
}

func getRegion(defaultRegion string) (string, error) {
	if len(defaultRegion) == 0 {
		return "", fmt.Errorf("default region is empty, please provide default region")
	}

	region, found := os.LookupEnv("CDK_DEPLOY_REGION")
	if !found {
		region, found = os.LookupEnv("CDK_DEFAULT_REGION")
		if !found {
			region, found = os.LookupEnv("AWS_REGION")
			if !found {
				region = defaultRegion
				fmt.Printf("no region found in CDK_DEPLOY_REGION, CDK_DEFAULT_REGION and AWS_REGION envs. Set %s as default\n", region)
			}
		}
	}
	return region, nil
}

func (cfg *Config) Log() {
	strByte, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal config struct for logging: %w", err))
	}
	fmt.Printf("config that is being used is \n %s", string(strByte))
}

type (
	Config struct {
		Env             string
		Account         string
		Region          string     `json:"region"`
		Stack           Stack      `json:"stack"`
		ApiGateway      ApiGateway `json:"apiGateway"`
		Lambda          Lambda     `json:"lambda"`
		Dynamo          Dynamo     `json:"dynamo"`
		ManagedPolicies []Policy   `json:"managedPolicies"`
	}

	Stack struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	ApiGateway struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Stage string `json:"stage"`
	}
	Lambda struct {
		ExecutionRole string `json:"executionRole"`
		Name          string `json:"name"`
	}

	Dynamo struct {
		Id           string `json:"id"`
		Name         string `json:"name"`
		PartitionKey string `json:"partitionKey"`
		SortKey      string `json:"sortKey"`
	}

	Policy struct {
		Name string `json:"name"`
		Arn  string `json:"arn"`
	}
)
