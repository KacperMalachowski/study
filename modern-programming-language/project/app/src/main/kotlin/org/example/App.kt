package org.example

import javafx.application.Application
import javafx.collections.FXCollections
import javafx.geometry.Insets
import javafx.scene.Scene
import javafx.scene.control.*
import javafx.scene.layout.GridPane
import javafx.scene.layout.HBox
import javafx.stage.Stage
import Task
import TaskRepository
import AddTaskUseCase
import UpdateTaskUseCase
import DeleteTaskUseCase
import GetAllTasksUseCase

class App : Application() {

    private val repository = TaskRepository
    private val addTaskUseCase = AddTaskUseCase(repository)
    private val updateTaskUseCase = UpdateTaskUseCase(repository)
    private val deleteTaskUseCase = DeleteTaskUseCase(repository)
    private val getAllTasksUseCase = GetAllTasksUseCase(repository)

    override fun start(primaryStage: Stage) {
        primaryStage.title = "TODO List Application"

        val grid = GridPane()
        grid.padding = Insets(10.0)
        grid.hgap = 10.0
        grid.vgap = 10.0

        val titleLabel = Label("Title")
        val titleField = TextField()
        val descriptionLabel = Label("Description")
        val descriptionField = TextArea()

        val addButton = Button("Add Task")
        val updateButton = Button("Update Task")
        val deleteButton = Button("Delete Task")

        val tasksList = ListView<Task>()
        tasksList.items = FXCollections.observableArrayList(getAllTasksUseCase.execute())

        grid.add(titleLabel, 0, 0)
        grid.add(titleField, 1, 0)
        grid.add(descriptionLabel, 4, 0)
        grid.add(descriptionField, 4, 2, 3, 2)
        grid.add(addButton, 0, 2)
        grid.add(updateButton, 1, 2)
        grid.add(deleteButton, 2, 2)
        grid.add(tasksList, 0, 3, 3, 1)

        addButton.setOnAction {
            val title = titleField.text
            val description = descriptionField.text
            if (title.isNotBlank() && description.isNotBlank()) {
                val task = addTaskUseCase.execute(title, description)
                tasksList.items.add(task)
                titleField.clear()
                descriptionField.clear()
            }
        }

        updateButton.setOnAction {
            val selectedTask = tasksList.selectionModel.selectedItem
            if (selectedTask != null) {
                selectedTask.title = titleField.text
                selectedTask.description = descriptionField.text
                updateTaskUseCase.execute(selectedTask)
                tasksList.refresh()
            }
        }

        deleteButton.setOnAction {
            val selectedTask = tasksList.selectionModel.selectedItem
            if (selectedTask != null) {
                deleteTaskUseCase.execute(selectedTask.id)
                tasksList.items.remove(selectedTask)
            }
        }

        tasksList.selectionModel.selectedItemProperty().addListener { _, _, selectedTask ->
            if (selectedTask != null) {
                titleField.text = selectedTask.title
                descriptionField.text = selectedTask.description
            }
        }

        val scene = Scene(grid, 600.0, 400.0)
        primaryStage.scene = scene
        primaryStage.show()
    }
}

fun main() {
    Application.launch(App::class.java)
}