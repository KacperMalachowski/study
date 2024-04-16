import numpy as np

def monte_carlo_square_aproximation(runs, a, b):
  xs = np.random.uniform(0, a, runs)
  ys = np.random.uniform(0, b, runs)

  inFigure = ((xs <= 2) & (ys <= xs)) | (((xs > 2) & (xs <= 6)) 
    & (ys <= -(xs/4)+2.5))  | ((xs > 6) & (ys <= -(xs/2)+4))

  mcArea = (sum(inFigure) / runs) * (a*b)

  return mcArea, xs, ys, inFigure