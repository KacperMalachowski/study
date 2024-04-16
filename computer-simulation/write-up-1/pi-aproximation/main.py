# import numpy as np
# import matplotlib.pyplot as plt

# runs = 100000
# xs = np.random.uniform(-0.5, 0.5, runs)
# ys = np.random.uniform(-0.5, 0.5, runs)
# in_circle = xs**2 + ys**2 <= 0.5**2
# mc_pi = (np.sum(in_circle) / runs) * 4

# plt.scatter(xs, ys, c=np.where(in_circle, 'blue', 'grey'), marker='.', edgecolors='none')
# plt.axis('equal')
# plt.title(f"MC Approximation of Pi = {mc_pi}")
# plt.xlabel('')
# plt.ylabel('')
# plt.show()

import numpy as np
import matplotlib.pyplot as plt

runs = 100000

xs = np.random.uniform(-0.5, 0.5, runs)
ys = np.random.uniform(-0.5, 0.5, runs)
inCircle = xs**2 + ys**2 <= 0.5**2
mcPi = (np.sum(inCircle) / runs) * 4

plt.scatter(xs, ys, c=np.where(inCircle, 'blue', 'grey'), marker='.', edgecolors='none')
plt.axis('equal')
plt.title("Zadanie 1 - Błąd aproksymacji")
plt.xlabel("Rozmiar próbki")
plt.ylabel("Błąd aproksymacji")
plt.show()