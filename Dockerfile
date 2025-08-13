FROM python:3.14.0rc-1-slim-bookworm

ARG USER=default
ARG GROUP=$USER
ARG PORT=5000

WORKDIR /macos-deployment

EXPOSE 5000