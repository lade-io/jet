try:
    from setuptools import setup
except ImportError:
    from distutils.core import setup

REQUIRES = ['numpy', 'pytest', 'flask']

config = {
    'description': 'Fibonacci webservice (with neural network)',
    'author': 'Soren Wacker',
    'url': 'https://github.com/soerendip/fibo',
    'download_url': 'https://github.com/soerendip/fibo',
    'author_email': 'swacker@posteo.de',
    'version': '0.1',
    'install_requires': [REQUIRES],
    'name': 'fibo',
    'packages': ['fibo'],
    'description': 'A webservice for fibonacci numbers.',
    'platform': 'Linux'
}

setup(**config)
