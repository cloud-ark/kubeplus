FROM python:slim
RUN pip3 install flask mysql-connector-python
COPY web-app.py /root/
COPY templates/ /root/templates
EXPOSE 5000
CMD ["python3", "/root/web-app.py"]
