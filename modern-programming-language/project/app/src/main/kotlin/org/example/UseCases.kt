
import TaskRepository

class AddTaskUseCase(private val repository: TaskRepository) {
    fun execute(title: String, description: String): Task {
        return repository.addTask(title, description)
    }
}

class UpdateTaskUseCase(private val repository: TaskRepository) {
    fun execute(task: Task) {
        repository.updateTask(task)
    }
}

class DeleteTaskUseCase(private val repository: TaskRepository) {
    fun execute(taskId: Int) {
        repository.deleteTask(taskId)
    }
}

class GetAllTasksUseCase(private val repository: TaskRepository) {
    fun execute(): List<Task> {
        return repository.getAllTasks()
    }
}