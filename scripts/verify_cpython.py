import json                                                                                                                        
import sys                                                                                                                       

HEADER_SIZE = sys.getsizeof([])    # 56 on 64-bit CPython
POINTER_SIZE = 8                    # sizeof(PyObject*) on 64-bit
                                                                                                                                    
def derive_allocated(lst):
    """Derive the internal 'allocated' field from sys.getsizeof."""                                                                
    return (sys.getsizeof(lst) - HEADER_SIZE) // POINTER_SIZE                                                                    
                                                                                                                                    
def main():
    lst = []                                                                                                                       
    results = []                                                                                                                 

    for i in range(1, 1001):
        lst.append(i)
        results.append({
            "append_num": i,
            "len": len(lst),
            "allocated": derive_allocated(lst),                                                                                    
        })
                                                                                                                                    
    json.dump(results, sys.stdout, indent=2)                                                                                     
    print()  # trailing newline

if __name__ == "__main__":
    main()