import matplotlib.pyplot as plt
import simulation as sim

runs = 100000
a = 8
b = 2

mc_area, xs, ys, in_figure = sim.monte_carlo_square_aproximation(runs, a, b)

plt.plot(xs[in_figure], ys[in_figure], '.', color='blue', label='Inside Area')
plt.plot(xs[~in_figure], ys[~in_figure], '.', color='grey', label='Outside Area')
plt.xlabel('')
plt.ylabel('')
plt.title("MC Approximation of area = {:.4f}".format(mc_area))
plt.gca().set_aspect('equal', adjustable='box')
plt.legend()
plt.savefig("../images/area_aproximation_single_run.png")

