package com.kacpermalachowski.ForEscape.models

import org.springframework.data.mongodb.core.mapping.Document
import org.springframework.data.annotation.Id
import org.bson.types.ObjectId

@Document("location")
data class Location(
  @Id
  val id: ObjectId = ObjectId(),
  val x: String = "",
  val y: String = ""
){}