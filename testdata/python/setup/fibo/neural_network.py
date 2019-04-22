import numpy as np
import tensorflow as tf


def singlelayer_perceptron(x, weights, n_input):
    '''
    Super simple network without bias terms.
    '''
    # Hidden layer with relu activation
    x = tf.transpose(x, [1, 0, 2])
    x = tf.reshape(x, [-1, n_input])
    # Convert to float32
    x = tf.cast(x, tf.float32)
    weights['h1'] = tf.cast(weights['h1'], tf.float32)
    layer_1 = tf.matmul(x, weights['h1'])
    layer_1 = tf.cast(tf.nn.relu(layer_1), tf.float32)
    # Output layer
    out_layer = (tf.matmul(layer_1, tf.cast(weights['out'], tf.float32)))
    return out_layer  # Parameters


def trained_network(seq):
    '''
    Trains a simple neural network based on the input sequence.
    It returns a function that can be used to make predictions based on the
    trained network. The networks is trained to predict the number that follows
    two adjacent numbers of the sequence.
        e.g.
            [0, 1, 1, 2, 3, 5, 8] --> predict(a, b)
            predict(3, 5) --> 8
    '''
    n = len(seq)
    # Network Parameters
    n_input = 2  # length of input sequence
    n_hidden = 2  # 1st layer number of features

    X = [seq[i-n_input:i] for i in range(n_input, n-n_input)]
    Y = [seq[i] for i in range(n_input, n-n_input)]
    print('Training data:', zip(X, Y))

    # Parameters
    learning_rate = 0.001
    total_batch = int(len(X))
    n_epoches = 2000

    # tf Graph input
    global x, y
    x = tf.placeholder("float32", [None, 1, n_input])
    y = tf.placeholder("float32", [None, None])

    # Store layers weight
    weights = {
        'h1': tf.Variable(tf.ones([n_input, n_hidden])),
        'out': tf.Variable(tf.random_normal([n_hidden, 1]))
    }

    # Construct model
    pred = singlelayer_perceptron(x, weights, n_input)

    # Define loss and optimizer
    # Using L2 loss function for prediction of real numbers
    cost = tf.nn.l2_loss(pred-y)
    optimizer = tf.train.AdamOptimizer(learning_rate=learning_rate).minimize(cost)

    # Initializing the variables
    init = tf.initialize_all_variables()

    # Launch the graph
    sess = tf.Session()
    sess.run(init)
    X_batches = np.array_split(X, total_batch)
    Y_batches = np.array_split(Y, total_batch)
    # Keep training until reach max iterations
    for epoche_i in range(n_epoches):
        for i in range(total_batch):
            batch_x, batch_y = [X_batches[i]], [Y_batches[i]]
            # Run optimization op (backprop)
            sess.run(optimizer, feed_dict={x: batch_x, y: batch_y})
    print("Optimization Finished!")
    print("The fitted data:")
    for X_i, Y_i in zip(X, Y):
        print(X_i, Y_i, sess.run(tf.cast(pred, tf.int32),
                                 feed_dict={x: [[X_i]]}))

    def predict(a, b, astype=float):
        '''
        Output of trained_model(). Contains a fitted tensorflow.session.
        Takes two values as input and applies network prediction.
        Returns float or type specified by astype argument.
            e.g.
                    predict(1, 2) --> astype(3)
        '''
        try:
            a = float(a)
        except:
            raise ValueError('First value could not be converted to float')
        try:
            b = float(b)
        except:
            raise ValueError('Second value could not be converted to float')
        print('Predicting', a, b, type(a), type(b))
        prediction = sess.run(pred, feed_dict={x: [[[a, b]]]})[0][0]
        print('Prediction:', prediction)
        return astype(prediction)

    return predict
