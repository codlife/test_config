FROM python:3.6
RUN mkdir -p /home/flask

WORKDIR /home/flask
COPY requirements.txt /home/flask
RUN pip3 install --no-cache-dir -i https://pypi.tuna.tsinghua.edu.cn/simple -r requirements.txt

COPY app /home/flask/

EXPOSE 8080

CMD ["python", "/home/flask/sentiment.py"]
