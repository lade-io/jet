def gen_first_n_fibonacci_numbers(n):
    '''
    Generator that returns the first n fibonacci numbers.
    Counting starts with 1.
        e.g.
              gen_first_n_fibonacci_numbers(5) --> 0, 1, 1, 2, 3
    '''
    try:
        float(n)
    except:
        raise ValueError('Value could not be converted to float')
    if not float(n).is_integer():
        raise ValueError('Value should be an integer')
    n = int(n)  # Convert to integer
    if n < 1:
        raise ValueError('Value must be >= 1')
    # Three cases to distinguish
    # n = 1, returns 0
    # n = 2, returns 0, 1
    # n > 2, returns 0, 1, then start iteration
    #        until n'th fibonacci number
    if n == 1:
        yield 0
    elif n == 2:
        yield 0
        yield 1
    else:
        yield 0
        yield 1
        a, b = 1, 0
        for i in range(n-2):
            a, b = a+b, a
            yield a


def fibonacci_sequence(n):
    '''
    Returns a list containing the first n fibonacci numbers.
        e.g.
            fiboacci_sequence(5) --> [0, 1, 1, 2, 3]
    '''
    return [i for i in gen_first_n_fibonacci_numbers(n)]
