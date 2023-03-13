# Kontroller
------------

## Description

This is a Kubernetes Admission Controller Webhook written in Go that contains two controller.
First one is `Validating Admission Webhook` that check a pod for `Annotations: validated-by: custom webhook` is present
or not and return response based on that. Second is the `Mutating Admission Webhook` that mutate
the pod and add `Annotations: validated-by: custom webhook` and return the response. By default, Kubernetes
call the Mutating webhook first then Validating one. Also, some key point to remember is that 
Kubernetes don't allow any webhook server that is not using any SSL/TLS certificate. So we have to 
generate SSL certificate as well. Please check [Makefile ssl](https://github.com/aniruddha2000/kontroller/blob/main/Makefile#L27-L31) 
command.

## Usage

### 1. Generate SSL certificate 

```shell
$ make ssl
```

This will generate the necessary SSL certificate in the `manifests/certs/` directory. Feel free to
change DNS name in the `manifests/certs/tls.cnf` file.

### 2. Start Kind Cluster

```shell
$ kind create cluster
```

### 3. Deploy manifests in the kind cluster

```shell
$ make manifest
```

### 4. Create a NGNIX pod

```shell
$ echo "apiVersion: v1
kind: Pod
metadata:
  name: webserver
spec:
  containers:
  - name: webserver
    image: nginx:latest
    ports:
    - containerPort: 80
  - name: webwatcher
    image: afakharany/watcher:latest" > ngnix.yaml | kubectl apply -f ngnix.yaml
```

Now describe the pod after creation you will see Annotations is being added. 