FROM python:3.10

RUN pip3 install pandas requests numpy vineyard scikit-learn==1.4.0 joblib==1.3.2

WORKDIR /

ARG APP
ENV APP=${APP}

COPY ${APP} /${APP}
