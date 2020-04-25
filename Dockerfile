ARG SPOTCTL_IMAGE=spotinst/spotctl:v0.0.12

ARG KUBECTL_VERSION=v1.18.2

FROM $SPOTCTL_IMAGE

ADD https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl /usr/local/bin/kubectl

RUN chmod +x /usr/local/bin/kubectl
