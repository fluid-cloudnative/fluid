# FROM python:3.10
FROM python:3.10@sha256:33f72df2ad8c9f777bf0adb35b9d89c5d62935cee2af1f9c3224fb6f7da1dc6b

# RUN pip3 install pandas requests numpy vineyard scikit-learn==1.4.0 joblib==1.3.2
COPY requirements.txt .
RUN pip3 install --no-cache-dir --require-hashes -r requirements.txt

WORKDIR /

ARG APP
ENV APP=${APP}

COPY ${APP} /${APP}
