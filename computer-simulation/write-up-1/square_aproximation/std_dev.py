import simulation as sim
import numpy as np

correct_value = 9
simulation_number = 100

runs = 100000
a = 8
b = 2

areas = np.zeros(simulation_number)

# Run n simulations
for simulation_id in range(simulation_number):
  areas[simulation_id], _, _, _ = sim.monte_carlo_square_aproximation(runs, a, b)

std_dev = np.std(areas)

print(f"Simulation count: {simulation_number}")
print(f"Standard deviation: {std_dev}")

