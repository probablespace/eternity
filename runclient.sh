#!/bin/bash

x=0
until nym/target/release/nym-client run --id etServe --gateway $@; do
	echo "nym client could not connect, retrying in 2 seconds.."
	sleep 2
done
