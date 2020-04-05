#!/usr/bin/env bash

wrk -t20 -c5000 -d10s -s ./loadtesting/randomise.lua --latency http://localhost:${PORT:=9010}
