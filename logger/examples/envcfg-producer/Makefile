all:	debug lru-kaf zap-kaf no-subj

debug:
	@PRLOG_AUTOINIT=true \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEDEBUG=true \
	go run main.go

lru-kaf:
	@PRLOG_AUTOINIT=true \
	PRLOG_LOGPACKAGE=logrus \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	go run main.go

zap-kaf:
	@PRLOG_AUTOINIT=true \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	go run main.go

no-subj:
	@PRLOG_AUTOINIT=true \
	PRLOG_ENABLECONSOLE=true \
	PRLOG_ENABLEFILE=false \
	PRLOG_ENABLEKAFKA=true \
	PRCE_SETSUBJECTLEVEL=false \
	go run main.go

test:
	@PRLOG_AUTOINIT=true \
	PRLOG_ENABLEKAFKA=true \
	PRKAFKA_TOPIC=test \
	PRCE_TYPE=com.github \
	PRROT_LOCALTIME=true \
	go run main.go
