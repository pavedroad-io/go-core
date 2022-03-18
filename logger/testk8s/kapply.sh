#!/usr/bin/env bash
# sample script to apply and expose kafka resources
kubectl apply -f zookeeper-deployment.yaml
kubectl apply -f zookeeper-service.yaml
kubectl apply -f kafka-broker-deployment.yaml
kubectl apply -f kafka-broker-service.yaml
kubectl wait deployment/zookeeper --for condition=available
kubectl wait deployment/kafka-broker --for condition=available
kubectl expose deployment kafka-broker --name=my-broker --external-ip=192.168.64.2
