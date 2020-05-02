[![Build Status](https://travis-ci.com/cprates/ippool.svg?token=xhTpgcEXoSMvxWuq6XB2&branch=master)](https://travis-ci.com/cprates/ippool)
[![Go Report Card](https://goreportcard.com/badge/github.com/cprates/ippool)](https://goreportcard.com/report/github.com/cprates/ippool)
[![License](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/cprates/lws/blob/master/LICENSE)


*ippool* is a dead simple IP pool to simulate the behavior of a DHCP server's ip pool.
It was designed to return sequentially generated IPs meant to have short lease periods, working
with a fairly small number of used IPs. Working with a almost exhausted pool will lead to longer
waiting times retrieving new IPs. 
