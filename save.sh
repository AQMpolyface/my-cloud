#!/bin/zsh

cp -r $flutter/transfer ./flutter-transfer/

cp -r $rust/transfer/ ./cli-transfer/



find . -type f -exec sed -i 's/my-cookie/my-cookie/g' {} + 
find . -type f -exec sed 's/polyface\.ch/example\.com/g' {} +

