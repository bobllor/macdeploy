FROM python:3.14-rc-slim-bookworm

ARG USER=python
ARG GROUP=$USER
ARG PORT=5000
ENV HOME=/home/${USER}/

RUN mkdir -p ${HOME}/.ca
RUN groupadd ${GROUP} && useradd ${USER} -g ${GROUP}

RUN apt-get install -y openssl
RUN openssl req -x509 newkey rsa:4096 -keyout ${HOME}/.ca/key.pem \
    -out ${HOME}/.ca/cert.pem -sha256 -days 3650 -node -subj "/CN=localhost"

WORKDIR /macos-deployment

COPY server/requirements.txt /tmp
RUN pip install -r /tmp/requirements.txt

USER ${USER}
EXPOSE ${PORT}

CMD ["gunicorn", "--workers=6", "--bind=0.0.0.0:5000", "--keyfile ${HOME}/.ca/key.pem", "--certfile ${HOME}/.ca/cert.pem", "server.app:app"]