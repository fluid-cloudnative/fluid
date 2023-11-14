# Fluid Documentation

This repository stores all the source files of Fluid documentation. That is, the documentation for [Fluid](https://github.com/fluid-cloudnative/fluid).

- `en`: [documentation in English](en/TOC.md)
- `zh`: [documentation in Chinese](zh/TOC.md)

## Documentation versions

Currently, we maintain the following versions for Fluid documentation, each with a separate branch:

| Branch name | Version description |
| :--- | :-- |
| `master` | the latest development version |
| `v0.1.0` | v0.1.0| 

## Generate a PDF Documentation
We also provide scripts that helps you generate a PDF documentation conveniently.
Before generating,we suppose you have installed [Docker](https://www.docker.com/) so you 
don't have to install required tools one by one.

1. Download Required Docker Image  
    `docker pull registry.cn-hangzhou.aliyuncs.com/docs-fluid/doc-build `
2. Start a Container  
    `docker run -it -v <your fluid/docs path>:/data/ fluid/doc-build:0.2.0`
3. Run Makefile
    ```shell
   cd data
   make build
   make clean
    ```