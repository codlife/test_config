FROM jupyter/base-notebook:7f1482f5a136
# Built from... https://hub.docker.com/r/jupyter/base-notebook/
#               https://github.com/jupyter/docker-stacks/blob/master/base-notebook/Dockerfile
# Built from... Ubuntu 18.04

# The jupyter/docker-stacks images contains jupyterhub, jupyterlab and the
# jupyterlab-hub extension already.

# Example install of git and nbgitpuller.
# NOTE: git is already available in the jupyter/minimal-notebook image.
USER root
RUN apt-get update && apt-get install --yes \
    git \
 && rm -rf /var/lib/apt/lists/*
USER $NB_USER

RUN if [ -d "~/.pip"]; then rm -r ~/.pip; fi
RUN mkdir -p ~/.pip && \
    echo "[global]" >> ~/.pip/pip.conf && \
    echo "index-url=http://mirrors.aliyun.com/pypi/simple/" >> ~/.pip/pip.conf && \
    echo "[install]" >> ~/.pip/pip.conf && \
    echo "trusted-host=mirrors.aliyun.com" >> ~/.pip/pip.conf && \
    pip install --no-cache-dir nbgitpuller && \
    jupyter serverextension enable --py nbgitpuller --sys-prefix

ADD requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir \
        -r /tmp/requirements.txt
#USER $NB_USER
RUN rm -r ~/.jupyter/* && \
    jupyter notebook -y --generate-config && \
    echo "c.NotebookApp.tornado_settings={'headers':{'Content-Security-Policy': \"frame-ancestors 'self' * \", 'Access-Control-Allow-Origin': '*'}}" >> ~/.jupyter/jupyter_notebook_config.py && \
    echo "c.NotebookApp.allow_origin='*'" >> ~/.jupyter/jupyter_notebook_config.py && \
    mkdir -p ~/.jupyter/custom && \
    echo "define(['base/js/namespace'], function(Jupyter) {Jupyter._target = '_self';});" > ~/.jupyter/custom/custom.js 

# Uncomment the line below to make nbgitpuller default to start up in JupyterLab
#ENV NBGITPULLER_APP=lab

# conda/pip/apt install additional packages here, if desired.
