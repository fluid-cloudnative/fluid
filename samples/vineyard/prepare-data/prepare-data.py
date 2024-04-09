import time

import numpy as np
import pandas as pd


def generate_random_dataframe(num_rows):
    return pd.DataFrame({
            'Id': np.random.randint(1, 100000, num_rows),
            'MSSubClass': np.random.randint(20, 201, size=num_rows),
            'LotFrontage': np.random.randint(50, 151, size=num_rows),
            'LotArea': np.random.randint(5000, 20001, size=num_rows),
            'OverallQual': np.random.randint(1, 11, size=num_rows),
            'OverallCond': np.random.randint(1, 11, size=num_rows),
            'YearBuilt': np.random.randint(1900, 2022, size=num_rows),
            'YearRemodAdd': np.random.randint(1900, 2022, size=num_rows),
            'MasVnrArea': np.random.randint(0, 1001, size=num_rows),
            'BsmtFinSF1': np.random.randint(0, 2001, size=num_rows),
            'BsmtFinSF2': np.random.randint(0, 1001, size=num_rows),
            'BsmtUnfSF': np.random.randint(0, 2001, size=num_rows),
            'TotalBsmtSF': np.random.randint(0, 3001, size=num_rows),
            '1stFlrSF': np.random.randint(500, 4001, size=num_rows),
            '2ndFlrSF': np.random.randint(0, 2001, size=num_rows),
            'LowQualFinSF': np.random.randint(0, 201, size=num_rows),
            'GrLivArea': np.random.randint(600, 5001, size=num_rows),
            'BsmtFullBath': np.random.randint(0, 4, size=num_rows),
            'BsmtHalfBath': np.random.randint(0, 3, size=num_rows),
            'FullBath': np.random.randint(0, 5, size=num_rows),
            'HalfBath': np.random.randint(0, 3, size=num_rows),
            'BedroomAbvGr': np.random.randint(0, 11, size=num_rows),
            'KitchenAbvGr': np.random.randint(0, 4, size=num_rows),
            'TotRmsAbvGrd': np.random.randint(0, 16, size=num_rows),
            'Fireplaces': np.random.randint(0, 4, size=num_rows),
            'GarageYrBlt': np.random.randint(1900, 2022, size=num_rows),
            'GarageCars': np.random.randint(0, 5, num_rows),
            'GarageArea': np.random.randint(0, 1001, num_rows),
            'WoodDeckSF': np.random.randint(0, 501, num_rows),
            'OpenPorchSF': np.random.randint(0, 301, num_rows),
            'EnclosedPorch': np.random.randint(0, 201, num_rows),
            '3SsnPorch': np.random.randint(0, 101, num_rows),
            'ScreenPorch': np.random.randint(0, 201, num_rows),
            'PoolArea': np.random.randint(0, 301, num_rows),
            'MiscVal': np.random.randint(0, 5001, num_rows),
            'TotalRooms': np.random.randint(2, 11, num_rows),
            "GarageAge": np.random.randint(1, 31, num_rows),
            "RemodAge": np.random.randint(1, 31, num_rows),
            "HouseAge": np.random.randint(1, 31, num_rows),
            "TotalBath": np.random.randint(1, 5, num_rows),
            "TotalPorchSF": np.random.randint(1, 1001, num_rows),
            "TotalSF": np.random.randint(1000, 6001, num_rows),
            "TotalArea": np.random.randint(1000, 6001, num_rows),
            'MoSold': np.random.randint(1, 13, num_rows),
            'YrSold': np.random.randint(2006, 2022, num_rows),
            'SalePrice': np.random.randint(50000, 800001, num_rows),
        })

def prepare_data():
    print('Start preparing data....', flush=True)
    st = time.time()
    for multiplier in 1000, 2000, 3000:
        df = generate_random_dataframe(10000*(multiplier))
        df.to_pickle('/data/df_{}.pkl'.format(multiplier))
        del df
    ed = time.time()
    print('##################################', flush=True)
    print('dataframe to_pickle time: ', ed - st, flush=True)


if __name__ == '__main__':
    st = time.time()
    print('Preparing data....', flush=True)
    prepare_data()
    ed = time.time()
    print('##################################')
    print('preparing data time: ', ed - st, flush=True)
    time.sleep(10000000)
