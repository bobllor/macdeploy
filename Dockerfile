FROM python:3.14-rc-slim-bookworm AS fsserver

ARG USER=pythonuser
ARG GROUP=${USER}
ARG PORT=5000
ENV HOME=/home/${USER}

RUN groupadd ${GROUP} && useradd ${USER} -g ${GROUP}
RUN mkdir ${HOME} && chown ${USER}:${GROUP} ${HOME}

RUN apt-get update && apt-get install -y openssl
SHELL [ "/bin/bash", "-c" ]

WORKDIR /macos-deployment

COPY requirements.txt ${HOME}

RUN pip install -r ${HOME}/requirements.txt

USER ${USER}
RUN mkdir -p ${HOME}/ca

RUN openssl req -x509 -newkey rsa:4096 -keyout ${HOME}/ca/key.pem \
    -out ${HOME}/ca/cert.pem -sha256 -days 3650 -nodes -subj "/CN=localhost"

EXPOSE ${PORT}

CMD /bin/bash -c "gunicorn --workers=6 \ 
--bind=0.0.0.0:5000 --keyfile $HOME/ca/key.pem \
--certfile $HOME/ca/cert.pem server.app:app"

##############################################################################
FROM golang:1.25.0-bookworm AS gopipe

ARG USER=gopuser
ARG GROUP=${USER}
ENV HOME=/home/${USER}
SHELL ["/bin/bash", "-c"]

RUN groupadd ${GROUP} && useradd ${USER} -g ${GROUP}
RUN mkdir ${HOME} && chown ${USER}:${GROUP} ${HOME}

WORKDIR /macos-deployment

##############################################################################
FROM debian:bookworm-slim AS cronner

ARG USER=cronuser
ARG GROUP=${USER}
ENV HOME=/home/${USER}
SHELL [ "/bin/bash", "-c" ]

RUN groupadd ${GROUP} && useradd ${USER} -g ${GROUP}
RUN mkdir ${HOME} && chown ${USER}:${GROUP} ${HOME}

RUN apt-get update \ 
    && apt-get install -y \
    cron curl procps

COPY scripts/zip_updater.sh ${HOME}

WORKDIR ${HOME}

RUN crontab -l | echo "*/10 * * * * ${HOME}/zip_updater.sh" > cron_stuff.txt \
    && crontab cron_stuff.txt

CMD ["cron", "-f"]