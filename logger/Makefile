.PHONY:	all env init pubs short long start clean

all:	long env init clean

env:	env-lru-con\
	env-lru-log\
	env-lru-pub\
	env-zap-con\
	env-zap-log\
	env-zap-pub

init:	init-lru-con\
	init-lru-log\
	init-lru-pub\
	init-zap-con\
	init-zap-log\
	init-zap-pub

pubs:	env-lru-pub\
	env-zap-pub\
	init-lru-pub\
	init-zap-pub

short:
	@go test -v -short

long:
	@go test -v

start:
	@docker-compose -f testdata/docker-compose.yaml up -d

clean:
	@docker ps | grep kafka >/dev/null && \
	docker-compose -f testdata/docker-compose.yaml down

.PHONY:	env-lru-con
env-lru-con:
	@PRTEST_ENV=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_CONSOLEWRITER=stdout \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /EnvLogrusConsoleDefault ${opts}

.PHONY:	env-lru-log
env-lru-log:
	@PRTEST_ENV=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_CONSOLEWRITER=stdout \
	PRLOG_ENABLEFILE=true \
	PRLOG_FILELOCATION=testdata/EnvLogrusLogfileDefault.log \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /EnvLogrusLogfileDefault ${opts}

.PHONY:	env-lru-pub
env-lru-pub: start
	@PRTEST_ENV=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	PRLOG_ENABLECLOUDEVENTS=true \
	go test -v -run /EnvLogrusPubsubDefault ${opts}

.PHONY:	env-zap-con
env-zap-con:
	@PRTEST_ENV=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=zap \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_CONSOLEWRITER=stdout \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /EnvZapConsoleDefault ${opts}

.PHONY:	env-zap-log
env-zap-log:
	@PRTEST_ENV=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=zap \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_CONSOLEWRITER=stdout \
	PRLOG_ENABLEFILE=true \
	PRLOG_FILELOCATION=testdata/EnvZapLogfileDefault.log \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /EnvZapLogfileDefault ${opts}

.PHONY:	env-zap-pub
env-zap-pub: start
	@PRTEST_ENV=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=zap \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	PRLOG_ENABLECLOUDEVENTS=true \
	go test -v -run /EnvZapPubsubDefault ${opts}

.PHONY:	init-lru-con
init-lru-con:
	@PRTEST_INIT=true \
	PRLOG_AUTOINIT=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_CONSOLEWRITER=testdata/InitLogrusConsoleDefault.log \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /InitLogrusConsoleDefault ${opts}

.PHONY:	init-lru-log
init-lru-log:
	@PRTEST_INIT=true \
	PRLOG_AUTOINIT=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_CONSOLEWRITER=stdout \
	PRLOG_ENABLEFILE=true \
	PRLOG_FILELOCATION=testdata/InitLogrusLogfileDefault.log \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /InitLogrusLogfileDefault ${opts}

.PHONY:	init-lru-pub
init-lru-pub: start
	@PRTEST_INIT=true \
	PRLOG_AUTOINIT=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	PRLOG_ENABLECLOUDEVENTS=true \
	go test -v -run /InitLogrusPubsubDefault ${opts}

.PHONY:	init-zap-con
init-zap-con:
	@PRTEST_INIT=true \
	PRLOG_AUTOINIT=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=zap \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_CONSOLEWRITER=testdata/InitZapConsoleDefault.log \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /InitZapConsoleDefault ${opts}

.PHONY:	init-zap-log
init-zap-log:
	@PRTEST_INIT=true \
	PRLOG_AUTOINIT=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=zap \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_CONSOLEWRITER=stdout \
	PRLOG_ENABLEFILE=true \
	PRLOG_FILELOCATION=testdata/InitZapLogfileDefault.log \
	PRLOG_ENABLEKAFKA=false \
	go test -v -run /InitZapLogfileDefault ${opts}

.PHONY:	init-zap-pub
init-zap-pub: start
	@PRTEST_INIT=true \
	PRLOG_AUTOINIT=true \
	PRLOG_CFGTYPE=env \
	PRLOG_LOGPACKAGE=zap \
	PRLOG_ENABLETIMESTAMPS=false \
	PRLOG_ENABLECONSOLE=false \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	PRLOG_ENABLECLOUDEVENTS=true \
	go test -v -run /InitZapPubsubDefault ${opts}
