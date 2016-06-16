#!/usr/bin/env bash

protoc -I ./com.iclick.adx/message/ ./com.iclick.adx/message/iclick_adx_rtb_v3.proto --go_out=plugins=grpc:com.iclick.adx/message