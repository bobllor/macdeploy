FROM python:3.14-rc-slim-bookworm

ARG USER=python
ARG GROUP=$USER
ARG PORT=5000
ENV HOME=/home/${USER}/

RUN mkdir -p ${HOME}/.ca
RUN groupadd ${GROUP} && useradd ${USER} -g ${GROUP}

RUN apt-get install -y openssl

WORKDIR /macos-deployment

COPY server/requirements.txt /tmp
RUN pip install -r /tmp/requirements.txt

USER ${USER}

EXPOSE ${PORT}

CMD ["gunicorn", "--workers=8", "--bind"]