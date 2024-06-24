import time

import numpy as np
import pandas as pd


def generate_random_dataframe(num_rows):
    rng = np.random.default_rng(seed=42)
    return pd.DataFrame({
            'Id': rng.integers(1, 100000, num_rows),
            'MSSubClass': rng.integers(20, 201, size=num_rows),
            'LotFrontage': rng.integers(50, 151, size=num_rows),
            'LotArea': rng.integers(5000, 20001, size=num_rows),
            'OverallQual': rng.integers(1, 11, size=num_rows),
            'OverallCond': rng.integers(1, 11, size=num_rows),
            'YearBuilt': rng.integers(1900, 2022, size=num_rows),
            'YearRemodAdd': rng.integers(1900, 2022, size=num_rows),
            'MasVnrArea': rng.integers(0, 1001, size=num_rows),
            'BsmtFinSF1': rng.integers(0, 2001, size=num_rows),
            'BsmtFinSF2': rng.integers(0, 1001, size=num_rows),
            'BsmtUnfSF': rng.integers(0, 2001, size=num_rows),
            'TotalBsmtSF': rng.integers(0, 3001, size=num_rows),
            '1stFlrSF': rng.integers(500, 4001, size=num_rows),
            '2ndFlrSF': rng.integers(0, 2001, size=num_rows),
            'LowQualFinSF': rng.integers(0, 201, size=num_rows),
            'GrLivArea': rng.integers(600, 5001, size=num_rows),
            'BsmtFullBath': rng.integers(0, 4, size=num_rows),
            'BsmtHalfBath': rng.integers(0, 3, size=num_rows),
            'FullBath': rng.integers(0, 5, size=num_rows),
            'HalfBath': rng.integers(0, 3, size=num_rows),
            'BedroomAbvGr': rng.integers(0, 11, size=num_rows),
            'KitchenAbvGr': rng.integers(0, 4, size=num_rows),
            'TotRmsAbvGrd': rng.integers(0, 16, size=num_rows),
            'Fireplaces': rng.integers(0, 4, size=num_rows),
            'GarageYrBlt': rng.integers(1900, 2022, size=num_rows),
            'GarageCars': rng.integers(0, 5, num_rows),
            'GarageArea': rng.integers(0, 1001, num_rows),
            'WoodDeckSF': rng.integers(0, 501, num_rows),
            'OpenPorchSF': rng.integers(0, 301, num_rows),
            'EnclosedPorch': rng.integers(0, 201, num_rows),
            '3SsnPorch': rng.integers(0, 101, num_rows),
            'ScreenPorch': rng.integers(0, 201, num_rows),
            'PoolArea': rng.integers(0, 301, num_rows),
            'MiscVal': rng.integers(0, 5001, num_rows),
            'TotalRooms': rng.integers(2, 11, num_rows),
            "GarageAge": rng.integers(1, 31, num_rows),
            "RemodAge": rng.integers(1, 31, num_rows),
            "HouseAge": rng.integers(1, 31, num_rows),
            "TotalBath": rng.integers(1, 5, num_rows),
            "TotalPorchSF": rng.integers(1, 1001, num_rows),
            "TotalSF": rng.integers(1000, 6001, num_rows),
            "TotalArea": rng.integers(1000, 6001, num_rows),
            'MoSold': rng.integers(1, 13, num_rows),
            'YrSold': rng.integers(2006, 2022, num_rows),
            'SalePrice': rng.integers(50000, 800001, num_rows),
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
