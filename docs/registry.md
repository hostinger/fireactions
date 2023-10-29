# Pull runner images from a private registry

nerdctl ghcr.io login -u <password> --password-stdin
nerdctl run -d -p 5000:5000 -e REGISTRY_PROXY_REMOTEURL=https://ghcr.io -e REGISTRY_PROXY_USERNAME="konradasb" -e REGISTRY_PROXY_PASSWORD="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" --restart always --name registry-ghcr.io registry:2
