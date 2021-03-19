#! /bin/sh

PS1='$ ' terminalizer

cat << EOS
echo "Hello World"
sleep 5
EOS | terminalizer record terminalizer.yaml
