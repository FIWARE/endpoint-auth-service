#!/bin/bash

echo "Analyze report"

failLevel=$1

low=$(cat clair-report.json | jq  ' .vulnerabilities[].normalized_severity | select(contains("Low"))' | wc -l)
medium=$(cat clair-report.json | jq  ' .vulnerabilities[].normalized_severity | select(contains("Medium"))' | wc -l)
high=$(cat clair-report.json | jq  ' .vulnerabilities[].normalized_severity | select(contains("High"))' | wc -l)
critical=$(cat clair-report.json | jq  ' .vulnerabilities[].normalized_severity | select(contains("Critical"))' | wc -l)

echo "CVE report: "
echo "Critical : $critical"
echo "High : $high"
echo "Medium : $medium"
echo "Low : $low"

if [ "$failLevel" = "low" ]; then
  if [ "$low" -gt "0" ]; then
    exit 1
  fi
elif [ "$failLevel" = "medium" ]; then
  if [ "$medium" -gt "0" ]; then
    exit 1
  fi
elif [ "$failLevel" = "high" ]; then
  if [ "$high" -gt "0" ]; then
    exit 1
  fi
elif [ "$failLevel" = "critical" ]; then
  if [ "$critical" -gt "0" ]; then
    exit 1
  fi
fi

exit 0