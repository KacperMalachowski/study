library(simmer)
library(parallel)

set.seed(42)

M1 <- 2 # Średni czas przetwarzania zapytań DNS
SD1 <- 2 #Odchylenie standardowe czasu przetwarzania zapytań DNS
M2 <- 4 # Średni czas przetwarzania zapytań HTML
SD2 <- 2 #Odchylenie standardowe czasu przetwarzania zapytań HTML
L_values <- seq(0.1, 1, 0.05)

run_simulation <- function(num_dns_html, num_html, lb_policy) {

  dns <- trajectory("dns_path") %>%
    select("dns_html", policy = lb_policy) %>% 
    seize_selected() %>%
    timeout(function() rnorm(1, M1, SD1)) %>%
    release_selected()

  html <- trajectory("html_path") %>%
    select(c("dns_html", "html"), policy = lb_policy) %>%
    seize_selected() %>%
    timeout(function() rnorm(1, M2, SD2)) %>%
    release_selected()

  envs <- mclapply(L_values, function(L) {
    q <- L 
    simmer("servers") %>%
      add_resource("dns_html", num_dns_html) %>%
      add_resource("html", num_html) %>%
      add_generator("dns", dns, function() rexp(1, q)) %>%
      add_generator("html", html, function() rexp(1, q)) %>%
      run(10000) %>%
      wrap()
  })

  results <- get_mon_arrivals(envs) %>% 
    transform(waiting_time = end_time - start_time - activity_time)

  mean_waiting_time <- c()
  for (i in 1:max(results$replication)){
    mean_waiting_time[i] <- mean(results$waiting_time[results$replication == i])
  }

  return(mean_waiting_time)
}

original_waiting_time <- run_simulation(3, 2, "shortest-queue")


plot(L_values, original_waiting_time, type="l", col="red",
      xlab="L (average requests per millisecond)", 
      ylab="mean waiting time [seconds]")

# one_html_failed_waiting_time <- run_simulation(3, 1, "shortest-queue")
# two_dns_html_failed_waiting_time <- run_simulation(1, 2, "shortest-queue")
# random_lb_waiting_time <- run_simulation(3, 2, "random")

# percentage_increase_one_html_failed <- (
#   (one_html_failed_waiting_time - original_waiting_time) / original_waiting_time
# ) * 100
# percentage_increase_two_dns_html_failed <- (
#   (two_dns_html_failed_waiting_time - original_waiting_time) / original_waiting_time
# ) * 100
# percentage_increase_random_lb <- (
#   (random_lb_waiting_time - original_waiting_time) / original_waiting_time
# ) * 100


# cat("Percentage increase in waiting time with one HTML server failure:", 
#     mean(percentage_increase_one_html_failed), "%\n")
# cat("Percentage increase in waiting time with two DNS+HTML server failures:", 
#     mean(percentage_increase_two_dns_html_failed), "%\n")
# cat("Percentage increase in waiting time with random load balancer:", mean(percentage_increase_random_lb), "%\n")