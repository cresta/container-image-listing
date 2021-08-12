# containerimagelisting
Library for listing difference attributes of images.

## Problem

Your docker images are in different container registries that all have their own API or authentication method, but you
need to list the tags for any of them.

## Solution

Use this library as a common way to list the container images of different container registries

## Example
```go

var opts RegistryFinderOptionalConfig
finder := RegistryFinder{
    Registries: []RegistryWithFinder{
        ForGHCR(cfg.GhcrUsername, cfg.GhcrPassword, opts),
        ForDockerhub(cfg.DockerhubUsername, cfg.DockerhubPassword, opts),
        ForQuay(cfg.QuayToken, opts),
        ForECR(ecr.New(ses), cfg.ECRBaseURL, opts),
    },
}

finder.ListTags("quay.io/bedrock/ubuntu")
finder.ListTags("ghcr.io/homebrew/brew")
finder.ListTags("ubuntu/redis")


```

## Local Testing

To test locally, run `mage go:test`