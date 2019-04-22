# Webservice to provide fibonacci numbers (with neural network)

To install tensorflow follow the instructions 
[here](https://www.tensorflow.org/versions/r0.9/get_started/os_setup.html#anaconda-installation).

I recommend the python2.7 version and using Anaconda to install tensorflow and than the package:

    conda create -n fibo python=2 ipython
    source activate fibo
    pip install --ignore-installed --upgrade https://storage.googleapis.com/tensorflow/linux/cpu/tensorflow-0.8.0-cp27-none-linux_x86_64.whl
    python setup.py install
or

    pip install fibo
    
    
After installation you can start the web-service with 

    run_service.py
or 

    run_network_service.py

Look at the output in the terminal to find out the IP address. Standard is http://127.0.0.1:5000/.


## run_service.py
Runs the simple server that returns the n'th number of the fibonacci sequence.

## run_network_service.py
Trains a small and simple neural network on triplets of the fibonacci sequence. The network was trained to predict the third number based on the first two numbers. Based on this information, the network approximates 'addition'. Due to the nature of the fibonacci sequence, which only contains positive numbers, the model only handles positive numbers. It has never seen a negative number, so it avoids predicting it.

