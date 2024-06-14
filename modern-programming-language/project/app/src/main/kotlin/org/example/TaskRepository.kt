object TaskRepository {
    private val tasks = mutableListOf<Task>()
    private var nextId = 1

    fun addTask(title: String, description: String): Task {
        val task = Task(nextId++, title, description)
        tasks.add(task)
        return task
    }

    fun updateTask(task: Task) {
        val index = tasks.indexOfFirst { it.id == task.id }
        if (index != -1) {
            tasks[index] = task
        }
    }

    fun deleteTask(taskId: Int) {
        tasks.removeAll { it.id == taskId }
    }

    fun getAllTasks(): List<Task> {
        return tasks.toList()
    }
}