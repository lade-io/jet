from flask import Flask
from flask import request
from flask import render_template
from fibo import fibonacci_sequence

app = Flask(__name__)


@app.route('/')
def my_form():
    return render_template("form.html")


@app.route('/', methods=['POST'])
def my_form_post():
    n = request.form['text']
    seq = fibonacci_sequence(n)
    return '<br>'.join([str(i) for i in seq])


@app.errorhandler(ValueError)
def handle_invalid_usage(error):
    return str(error)

if __name__ == "__main__":
    app.run()
