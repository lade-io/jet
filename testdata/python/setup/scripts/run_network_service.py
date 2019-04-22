from flask import Flask
from flask import request
from flask import render_template
from fibo import fibonacci_sequence
from fibo.neural_network import trained_network


app = Flask(__name__)


@app.route('/')
def my_form():
    return render_template("form_network_service.html")


@app.route('/', methods=['POST'])
def my_form_post():
    A = request.form['A']
    B = request.form['B']
    prediction = predict(A, B, astype=str)
    return prediction


@app.errorhandler(ValueError)
def handle_invalid_usage(error):
    return str(error)


if __name__ == "__main__":
    predict = trained_network(fibonacci_sequence(10))
    app.run()
