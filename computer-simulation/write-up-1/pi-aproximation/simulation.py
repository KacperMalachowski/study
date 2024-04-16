import numpy as np

def monte_carlo_pi_aproximation(runs):
  xs = np.random.uniform(-0.5, 0.5, runs)
  ys = np.random.uniform(-0.5, 0.5, runs)
  inCircle = xs**2 + ys**2 <= 0.5**2

  mcPi = (np.sum(inCircle) / runs) * 4

  return mcPi