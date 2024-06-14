import kotlin.test.*

class TaskRepositoryTest {
    private lateinit var repository: TaskRepository

    @BeforeTest
    fun setUp() {
        repository = TaskRepository
        // Clear the repository before each test
        repository.getAllTasks().forEach { repository.deleteTask(it.id) }
    }

    @Test
    fun testAddTask() {
        val task = repository.addTask("Test Task", "Test Description")
        assertEquals(1, repository.getAllTasks().size)
        assertEquals("Test Task", task.title)
        assertEquals("Test Description", task.description)
    }

    @Test
    fun testUpdateTask() {
        val task = repository.addTask("Test Task", "Test Description")
        task.title = "Updated Task"
        task.description = "Updated Description"
        repository.updateTask(task)

        val updatedTask = repository.getAllTasks().first()
        assertEquals("Updated Task", updatedTask.title)
        assertEquals("Updated Description", updatedTask.description)
    }

    @Test
    fun testDeleteTask() {
        val task = repository.addTask("Test Task", "Test Description")
        repository.deleteTask(task.id)
        assertTrue(repository.getAllTasks().isEmpty())
    }

    @Test
    fun testGetAllTasks() {
        repository.addTask("Task 1", "Description 1")
        repository.addTask("Task 2", "Description 2")

        val tasks = repository.getAllTasks()
        assertEquals(2, tasks.size)
    }
  }
