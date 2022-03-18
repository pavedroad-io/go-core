#!/usr/bin/env bash
# sample script to delete kafka resources
kubectl delete -f kafka-broker-service.yaml
kubectl delete -f kafka-broker-deployment.yaml
kubectl delete -f zookeeper-service.yaml
kubectl delete -f zookeeper-deployment.yaml
kubectl delete service my-broker
kubectl wait --for=delete service/my-broker
kubectl wait --for=delete deployment/Kafka-broker
kubectl wait --for=delete deployment/zookeeper
