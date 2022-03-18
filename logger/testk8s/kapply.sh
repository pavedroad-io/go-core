#!/usr/bin/env bash
# sample script to apply and expose kafka resources
kubectl apply -f zookeeper-deployment.yaml
kubectl apply -f zookeeper-service.yaml
kubectl wait --for condition=available deployment/zookeeper
kubectl apply -f kafka-broker-deployment.yaml
kubectl apply -f kafka-broker-service.yaml
kubectl wait --for condition=available deployment/kafka-broker
kubectl expose deployment kafka-broker --name=my-broker --external-ip=192.168.64.2
