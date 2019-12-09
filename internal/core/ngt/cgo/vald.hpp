#ifndef __VALD_H__
#define __VALD_H__

#include <NGT/Capi.h>
#include <NGT/Index.h>
#include <stdlib.h>
#include <iostream>

#ifdef __cplusplus
extern "C" {
#endif

static bool operate_error_string_(const std::stringstream &ss, NGTError error){
  if(error != NULL){  
    try{
      std::string *error_str = static_cast<std::string*>(error);
      *error_str = ss.str();
    }catch(std::exception &err){
      std::cerr << ss.str() << " > " << err.what() << std::endl;
      return false;
    }
  }else{
    std::cerr << ss.str() << std::endl;
  }
  return true;
}

bool ngt_bulk_insert_index(NGTIndex index, float *obj, uint32_t data_count, uint32_t *ids, NGTError error) {
    std::cout << obj[17] << std::endl;
  NGT::Index* pindex = static_cast<NGT::Index*>(index);
  int32_t dim = pindex->getObjectSpace().getDimension();

  bool status = true;
  float *objptr = obj;
  for (size_t idx = 0; idx < data_count; idx++, objptr += dim) {
    try{
      std::vector<double> vobj(objptr, objptr + dim);
      ids[idx] = pindex->insert(vobj);
    }catch(std::exception &err) {
      status = false;
      ids[idx] = 0;
      std::stringstream ss;
      ss << "Capi : " << __FUNCTION__ << "() : Error: " << err.what();
      operate_error_string_(ss, error);      
    }
  }
  return status;
}


#ifdef __cplusplus
}
#endif

#endif
