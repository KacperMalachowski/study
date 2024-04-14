import pandas as pd
import numpy as np

df = pd.read_csv("./Occupancy_Estimation.csv")

column_name1 = 'S4_Sound'
column_name2 = 'S5_CO2_Slope'

num_values_to_replace = int(0.05 * len(df))

replace_indices1 = np.random.choice(df.index, num_values_to_replace, replace=False)
replace_indices2 = np.random.choice(df.index, num_values_to_replace, replace=False)

def new_value_generator():
  return "?"

df.loc[replace_indices1, column_name1] = new_value_generator()
df.loc[replace_indices2, column_name2] = new_value_generator()

df.to_csv('modified_file.csv', index=False)

replaced_rows_count = df[(df[column_name1] == "?") | (df[column_name2] == "?")].shape[0]

# Calculate the percentage of rows with replaced values
percentage_replaced_rows = (replaced_rows_count / len(df)) * 100

print(f"Percentage of rows with replaced data: {percentage_replaced_rows:.2f}%")