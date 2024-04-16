import numpy as np
import matplotlib.pyplot as plt
import simulation as sim

min_runs = 100000
increase = 10000
max_runs = 1000000
pi_value = 3.14159265

results = {}

for runs in range(min_runs, max_runs, increase):
  mc_pi = sim.monte_carlo_pi_aproximation(runs)
  results[runs] = abs(mc_pi - pi_value)

linear_regression = np.polyfit(list(results.keys()), list(results.values()), 1)

plt.scatter(results.keys(), results.values())
plt.plot(results.keys(), np.polyval(linear_regression, list(results.keys())), color="red")
plt.title("Zadanie 1 - Błąd aproksymacji")
plt.xlabel("Rozmiar próbki")
plt.ylabel("Błąd aproksymacji")
plt.savefig("../images/pi_aproximation_based_on_sample_size.png")