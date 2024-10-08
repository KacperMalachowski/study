package com.kacpermalachowski.ForEscape.repositories

import org.springframework.data.mongodb.repository.MongoRepository
import org.springframework.stereotype.Repository
import com.kacpermalachowski.ForEscape.models.Location

@Repository
interface LocationRepository : MongoRepository<Location, String>{
  
}