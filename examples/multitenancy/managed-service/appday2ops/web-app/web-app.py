import os
from flask import Flask, render_template, request
import mysql.connector
from mysql.connector import Error

app = Flask(__name__)

def create_mysql_connection():
    connection = None
    try:
        connection = mysql.connector.connect(
            host=os.getenv('MYSQL_HOST'),
            user=os.getenv('MYSQL_USER'),
            password=os.getenv('MYSQL_PASSWORD'),
            database=os.getenv('MYSQL_DATABASE'),
        )
        print("MySQL connection successful.")
    except Error as e:
        print(f"Error: '{e}'")
    return connection

def get_users():
    connection = create_mysql_connection()
    if not connection:
        return []
    cursor = connection.cursor()
    cursor.execute("SELECT * FROM users")
    userList = cursor.fetchall()
    userList = [user[0] for user in userList]
    connection.close()
    return userList

def create_user(username):
    connection = create_mysql_connection()
    if not connection:
        return False
    cursor = connection.cursor()
    cursor.execute(f"INSERT INTO users (name) VALUES ('{username}')")
    connection.commit()
    connection.close()
    return True

@app.route("/", methods=["GET", "POST"])
def index():
    if request.method == "POST":
        username = request.form["username"]
        if username in get_users():
            return f"{username} is already in the database"
        else:
            if create_user(username):
                return f"{username} added to database"
            else:
                return f"Couldn't add {username} to the database"
    return render_template("form.html")

@app.route("/users")
def users():
    userList = get_users()
    if userList:
        return render_template("users.html", userList=userList)
    return "No users"

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)

