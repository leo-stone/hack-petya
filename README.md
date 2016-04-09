# hack-petya mission accomplished!!!
============

## find key in seconds to restore petya ransomware encrypted mft

My easter visit to my father in law got me into this mess, _excuse me_.
First news after the hello was: 
>"Someone was applying for a job, and i really, really needed to read his 
>CV, so I entered the ADMIN-PASSWORD and now there is only this red skull ..."

*oh no, oh nooo*

Well, I always like a challenge ..., the hard task of analyzing and reimplementing the modified salsa algorithm is done.  
So, here it is for everyone to play and experiment with. _Btw. paying ransom isn't that much of a challenge_.

The code reimplements the hashing used to verify the entered key, there is still the possibility, that the real decryption uses a different hashing function. But at the first glance it seemed the same. I hadn't had the time to look deeper yet.

### Some key points: 

* its salsa, yes, but it operates on 16-bit words, not 32-bit

* its not salsa**20**, but salsa**10**, e.g. it shuffles the matrix only for ten rounds

Data Locations:  
  * Nonce 8-bytes: 
     - sector 54 [0x36] offset: 33 [0x21] 
  * Encrypted Verification Sector 512-bytes:   
     - sector 55 [0x37] offset: 0 [0x0] 

The code reads a file "src.txt" which is the 512-bytes from sector 55, pulls the interesting words, xor's them with 0x37 to give us the target output of the salsa hashing function.

It also reads "nonce.txt" which is the 8-byte nonce that was used in the attack.

Last, but not least, it fires up a genetic solver which gets us the key in a few seconds.

I recovered my key in say 10..30 seconds :), i just say Genetic Algorithms

*PS: I know the code is a mess, but is was kinda in a hurry ..., i also had to hack into the genetic lib 'cause its not
compatible with go1.6, concurrent map read/writes panics.*




