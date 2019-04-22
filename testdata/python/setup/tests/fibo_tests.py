import pytest
from fibo import gen_first_n_fibonacci_numbers, fibonacci_sequence


def test_fibonacci_sequence__negative_raises_ValueError():
    with pytest.raises(ValueError):
        fibonacci_sequence(-1)


def test_fibonacci_sequence__len_is_equal_to_n():
    for n in [1, 2, 3, 4, 5, 100]:
        assert len(fibonacci_sequence(n)) == n,\
            'length of output %d != %d' % (len(fibonacci_sequence(n)), n)


def test_fibonacci_sequence__large_int_works():
    n = 100000
    assert len(fibonacci_sequence(n)) == n,\
        'length of output %d != %d' % (len(fibonacci_sequence(n)), n)


def test_fibonacci_sequence__non_float_raises_ValueError():
    with pytest.raises(ValueError):
        fibonacci_sequence(1.1)


def test_fibonacci_sequence__zero_raises_ValueError():
    with pytest.raises(ValueError):
        fibonacci_sequence(0)


def test_fibonacci_sequence__accepts_string():
    assert fibonacci_sequence('50') == fibonacci_sequence(50)


def test_gen_first_n_fibonacci_numbers__raises_ValueError():
    with pytest.raises(ValueError):
        for i in gen_first_n_fibonacci_numbers('a'):
            pass
