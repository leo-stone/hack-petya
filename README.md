# hack-petya
============

## Search key to restore petya ransomware encrypted mft

My easter visit to my father in law got me into this mess, _excuse me_.
First news after the hello was: 
>"Someone was applying for a job, and i really, really needed to read his 
>CV, so I entered the ADMIN-PASSWORD and now there is only this red skull ..."

*oh no, oh nooo*

Well, i always like a challenge ..., the hard task of analyzing and reimplementing the modified salsa algorithm is done.
So, here it is for everyone to play and experiment with. _Btw. paying ransom isn't that much of a challenge_.

The code reimplements the hashing used to verify the entered key, there is still the possibility, that the real decryption uses a different hashing function. But at the first glance it seemed the same. I hadn't had the time to look deeper yet.

### Some key points: 

* its salsa, yes, but it operates on 16-bit words, not 32-bit

* its not salsa**20**, but salsa**10**, e.g. it shuffles the matrix only for ten rounds

I might list some important disk positions later, until now i was just following the execution, and grabbed the needed data on the fly
from memory.

For now I only remember sector 55 / 0x37 which is the 512-byte data (256 bytes thereof) used to verify the key.

The position of the nonce i still have to look up.

The code reads a file "src.txt" which is the 512-bytes from sector 55, pulls the interesting words, xor's them with 0x37 to give us the target output of the salsa hashing function.

It also reads "nonce.txt" which is the 8-byte nonce that was used in the attack.

Last, but not least, it fires up 24 threads to search the keyspace via bruteforce for a match.

### ToDo:

Possibly find a better strategy than bruteforce ...... 




