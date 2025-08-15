FROM python:3.14-rc-slim-bookworm

ARG USER=python
ARG GROUP=$USER
ARG PORT=5000
ENV HOME=/home/${USER}/

RUN groupadd ${GROUP} && useradd ${USER} -g ${GROUP}
RUN mkdir ${HOME} && chown ${USER}:${GROUP} ${HOME}

RUN apt-get install -y openssl
SHELL [ "/bin/bash", "-c" ]

WORKDIR /macos-deployment

COPY server/requirements.txt ${HOME}

RUN pip install -r ${HOME}/requirements.txt

USER ${USER}
RUN mkdir -p ${HOME}/ca

RUN openssl req -x509 -newkey rsa:4096 -keyout ${HOME}/ca/key.pem \
    -out ${HOME}/ca/cert.pem -sha256 -days 3650 -nodes -subj "/CN=localhost"

EXPOSE ${PORT}

CMD ["gunicorn", "--workers=6", \ 
"--bind=0.0.0.0:5000", "--keyfile", "/home/python/ca/key.pem", \
"--certfile", "/home/python/ca/cert.pem", "server.app:app"]