package cfgx

// GetConfig returns the configuration as found in the meta config
//
//	By default, the lookup logic is as follows:
//	    a. Use the path specified on the command line via --meta-config
//	    b. Use the path specified in the environment variable: kenv.meta-config
//	    c. Look in the current application's working directory for a file called
//	         meta-config.yaml or meta-config.yml
//
//	Inside the resolved meta config, a top level section called 'env-settings'
//	is used to inject key/value pairs into the environment variables via os.Setenv().
//	This is done prior to parsing config files specified in the meta config.
//	All config files resolved are then updated using os.ExpandEnv(). This will
//	inject (e.g., replace) any environment variable placeholders (subject to
//	the same rules indicated for os.ExpandEnv()). When specifying a file source,
//	you can optionally specify a selection path to select a subset of the config
//	file referenced. See 'messaging-emitter' below for an example:
//
//	## Example meta-config.yaml ##
//	env-settings: "~/.kwil-server/env-settings.yaml"
//	tracking-service: "./tracking-service.yaml"
//	messaging-emitter: "messaging-emitter, ../mx/kafka-topic-settings.yaml"
//
//
//	## And inside env-settings.yaml ##
//	KAFKA_CLUSTER_CONSUMER:
//	    ENDPOINT: pkc-2396y.us-east-1.aws.confluent.cloud:9092
//	    USER: KWIL_CONSUMER
//	    PASSWORD: 1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ-etc <- NOTE: example only
//
//	DB_CLUSTER:
//	    ENDPOINT: kwil-dev-cluster-instance-1.abcdefghhiklmnop.us-east-1.rds.amazonaws.com
//	    USER: kwil_rw
//	    PASSWORD: ABCDEFGHIJKLMNOP123! <- NOTE: example only
func GetConfig() Config {
	return getConfigInternal()
}

func GetTestConfig() Config {
	return getTestConfigInternal()
}
