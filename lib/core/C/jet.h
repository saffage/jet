#ifndef _JET_LANGUAGE_H_
#define _JET_LANGUAGE_H_

//
// built-in types
//

#include <stdint.h>

typedef int8_t    jet_i8;
typedef int16_t   jet_i16;
typedef int32_t   jet_i32;
typedef int64_t   jet_i64;
typedef uint8_t   jet_u8;
typedef uint16_t  jet_u16;
typedef uint32_t  jet_u32;
typedef uint64_t  jet_u64;
typedef intptr_t  jet_isize;
typedef uintptr_t jet_usize;
typedef float     jet_f32;
typedef double    jet_f64;
typedef jet_u8    jet_char;
typedef jet_u32   jet_rune;

//
// core types
//

typedef struct jet_slice   jet_slice;
typedef struct jet_seq     jet_seq;
typedef struct jet_string  jet_string;

struct jet_slice {
  void*     data;
  jet_usize len;
};
struct jet_seq {
  void*     data;
  jet_usize len;
  jet_usize cap;
};
struct jet_string {
  jet_u8*   data;
  jet_usize len;
  jet_usize cap;
};

#endif // _JET_LANGUAGE_H_
