package com.kacpermalachowski.ForEscape.controllers

import com.kacpermalachowski.ForEscape.repositories.LocationRepository
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.DeleteMapping
import org.springframework.web.bind.annotation.RequestBody
import com.kacpermalachowski.ForEscape.models.Location;

@RestController
@RequestMapping("/locations")
class LocationController(val locationRepository: LocationRepository){

  @GetMapping()
  fun getAllLocations() = locationRepository.findAll()

  @PostMapping()
  fun addNewLocation(@RequestBody location: Location) = locationRepository.insert(location)

  @DeleteMapping()
  fun deleteLocation(@RequestBody location: Location) = locationRepository.delete(location)
}
