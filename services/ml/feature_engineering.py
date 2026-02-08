import pandas as pd
import numpy as np
from sklearn.preprocessing import RobustScaler
import hashlib
import logging

logger = logging.getLogger(__name__)

class FeatureEngineer:
    def __init__(self):
        self.scaler = RobustScaler() 
        self.feature_names = []
        self.is_fitted = False

    def extract_features(self,df:pd.DataFrame) ->pd.DataFrame:
        if df.empty : return df

        features_df = df.copy()

        if 'timestamps' in df.columns:
            ts = pd.to_datetime(df['timestamp'])
            features_df['hour'] = ts.dt.hour
            features_df['day_of_week'] = ts.dt.dayofweek
            features_df['hour_sin'] = np.sin(2 * np.pi * features_df['hour'] / 24)
            features_df['hour_cos'] = np.cos(2 * np.pi * features_df['hour'] / 24)

            if 'method' in df.columns:
                method_map = {'GET':0 , 'POST':1,'PUT':2,'DELETE':3,'PATCH':4}
                features_df['method_encoded'] = df[method_map].fillna(5)
            
            if 'path' in df.columns:
                features_df['path_length'] = df['path'].str.len()
                features_df['path_depth'] = df['path'].str.count('/')
                features_df['has_query'] = df['path'].str.contains(r'\?').astype(int)

                suspicious = ['select', 'union', 'exec', '../', '/etc/passwd']
            features_df['is_suspicious_path'] = df['path'].str.lower().apply(
                lambda x: any(s in str(x) for s in suspicious)
            ).astype(int)
            
        # 3. Boyut ve Yanıt Süresi
        for col in ['response_time', 'request_size', 'response_size']:
            if col in df.columns:
                features_df[f'{col}_log'] = np.log1p(df[col].astype(float))
        
        # 4. Durum Kodları
        if 'status_code' in df.columns:
            features_df['is_error'] = (df['status_code'] >= 400).astype(int)
            features_df['is_server_error'] = (df['status_code'] >= 500).astype(int)
            
       
        numeric_features = features_df.select_dtypes(include=[np.number]).columns.tolist()
        
        numeric_features = [c for c in numeric_features if c not in ['_id']]
        
        self.feature_names = numeric_features
        result_df = features_df[numeric_features].fillna(0)
        
        return result_df

    def transform(self, features_df: pd.DataFrame) -> np.ndarray:
        """Veriyi [0, 1] veya [-1, 1] arasına ölçekler (Modelin daha iyi öğrenmesi için)"""
        if not self.is_fitted:
            self.scaler.fit(features_df)
            self.is_fitted = True
        return self.scaler.transform(features_df)

if __name__ == '__main__':
    print("Feature Engineer sınıfı hazır.")