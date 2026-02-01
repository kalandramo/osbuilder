{{- $D := or .Web .MQ .Job -}}
#!/bin/bash

kubectl create configmap {{$D.BinaryName}} --from-file=../../{{$D.BinaryName}}.yaml --dry-run=client -o yaml|kubectl apply -f  -
