data class Task(
    val id: Int,
    var title: String,
    var description: String
) {
  override fun toString(): String = "$id $title"
}