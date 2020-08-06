import os
import time
import sys
from glob import glob
from threading import Thread

from six.moves.queue import Queue
import argparse

def get_file_recursively(db_glob):
    db = glob(db_glob)
    res = []
    for f in db:
        if os.path.isdir(f):
            res.extend(get_file_recursively("{}/*".format(f)))
        else:
            res.append(f)
    
    return res

def worker(db, index_queue, result_queue):
    while True:
        ind = index_queue.get()
        if ind is None:
            return
        image_path = db[ind]
        if os.path.isdir(image_path):
            result_queue.put(0)
            return
        fd = open(image_path, 'rb')
        im = fd.read()
        fd.close()
        result_queue.put(len(im))


def test_db(db_glob, num_thread):
    # python 2.7 does not support glob such as ".../**" to do recursively
    # to access all files under some directory, we have to implement ourselves.
    db = get_file_recursively(db_glob)

    # using python 3.5, a simple form:
    # db = glob(db_glob)

    print("{} contains {} items".format(db_glob, len(db)))

    index_queue = Queue()
    result_queue = Queue()                                
    workers = [Thread(target=worker, args=(db, index_queue, result_queue)) for _
                      in range(num_thread)]                                     
    for w in workers:                                                           
        w.start()                                                               
                                                                                
    t1 = time.time()                                                            
    for i in range(len(db)):                                                    
        index_queue.put(i)                                                      
    total_len = 0                                                               
    for i in range(len(db)):                                                    
        total_len += result_queue.get()                                         
    t2 = time.time() 

    for i in range(len(workers)):
        index_queue.put(None)
    
    print("{} processing {} items with {} threads uses {}s, avg {}/s, avg {}B/s, avg {}MiB/s".format(db_glob, len(db), num_thread, (t2 - t1),
                                                            len(db) / (t2 - t1), total_len / (t2 - t1), total_len / (t2 - t1) / 1024 / 1024))

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--threads', type=int, default=2, help='The number of threads')
    parser.add_argument('--path', type=str, help='Data path')
    FLAGS, unparsed = parser.parse_known_args()
    test_db(FLAGS.path, FLAGS.threads)
