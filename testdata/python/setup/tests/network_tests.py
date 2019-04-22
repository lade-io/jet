import pytest
from fibo.neural_network import trained_network


def test_trained_network__generates_predict_func():
    predict = trained_network([1, 2, 3, 4, 5])
    assert isinstance(predict(1, 2), float)
    assert isinstance(predict(1, 2, astype=str), str)
    assert isinstance(predict(1, 2, astype=int), int)


def test_trained_network__raises_ValueErrors():
    predict = trained_network([1, 2, 3, 4, 5])
    with pytest.raises(ValueError):
        predict('a', 1)
    with pytest.raises(ValueError):
        predict(1, 'b')
