#!/bin/bash
# Example telnet interaction with redis-go

(
  sleep 1
  printf "SET foo bar\r\n"
  sleep 1
  printf "GET foo\r\n"
  sleep 1
  printf "EXPIRE foo 3\r\n"
  sleep 1
  printf "TTL foo\r\n"
  sleep 4
  printf "GET foo\r\n"
  sleep 1
  printf "QUIT\r\n"
) | telnet localhost 6379
